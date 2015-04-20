#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <pthread.h>
#include <semaphore.h>
#include <stdint.h>
#include <errno.h>
#include <unistd.h>
#include <sys/time.h>
#include <time.h>
#include "libusb.h"
#include "sqlite3.h"

//------------------------------------------------------------------------------
// Defines 
//------------------------------------------------------------------------------
#ifndef ID_VENDOR
#define ID_VENDOR        0x10C4
#endif
#ifndef ID_PRODUCT
#define ID_PRODUCT        0x8A40
#endif
#ifndef EP_ADDRESS
#define EP_ADDRESS        0x83
#endif
#ifndef WAL_CHECKOUT_LENGTH
#define WAL_CHECKOUT_LENGTH 1000
#endif
//------------------------------------------------------------------------------
// Firmware
//------------------------------------------------------------------------------
#define FIRMWARE_ERROR             -1

#define FIRMWARE_CA_TEST           0
#define FIRMWARE_CA_TEST__MF       "Silicon Laboratories Inc."
#define FIRMWARE_CA_TEST__MF_LEN   sizeof(FIRMWARE_CA_TEST__MF)
#define FIRMWARE_CA_TEST__PD       "Fake Streaming 64byt"
#define FIRMWARE_CA_TEST__PD_LEN   sizeof(FIRMWARE_CA_TEST__PD)

#define FIRMWARE_CA_PULSE_OXIMETER         1
#define FIRMWARE_CA_PULSE_OXIMETER__MF     "CardioArt"
#define FIRMWARE_CA_PULSE_OXIMETER__MF_LEN sizeof(FIRMWARE_CA_PULSE_OXIMETER__MF) -1
#define FIRMWARE_CA_PULSE_OXIMETER__PD     "Pulse Oximeter"
#define FIRMWARE_CA_PULSE_OXIMETER__PD_LEN sizeof(FIRMWARE_CA_PULSE_OXIMETER__PD) -1

#define FIRMWARE_CA_ECG_MONITOR            2
#define FIRMWARE_CA_ECG_MONITOR__MF        "CardioArt"
#define FIRMWARE_CA_ECG_MONITOR__MF_LEN    sizeof(FIRMWARE_CA_ECG_MONITOR__MF)-1
#define FIRMWARE_CA_ECG_MONITOR__PD        "ECG Monitor"
#define FIRMWARE_CA_ECG_MONITOR__PD_LEN    sizeof(FIRMWARE_CA_ECG_MONITOR__PD)-1
//------------------------------------------------------------------------------
// Macros
//------------------------------------------------------------------------------
// macro to handle ERROR
#define TRY_LIBUSB(msg, ret_err) \
    do { if (ret_err != 0) { \
    printf("error %s:%s\n", msg, libusb_error_name(ret_err)); goto exit_main;} \
    } while(0)
#define TRY_LIBUSB2(msg, ret_err) \
    do { if (ret_err != 0) { \
    printf("error %s\n", msg);}} while(0)
#define TRY(msg, ret_err) \
    do { if (ret_err != 0) { \
    printf("error %s\n", msg);}} while(0)
#define TRY_STOP(msg, ret_err) \
    do { if (ret_err != 0) { \
    printf("error %s\n", msg); goto exit_main;}} while(0)
#define TRY_SQLITE(msg, ret_err) \
    do { if (ret_err != SQLITE_OK) { printf("error %s\n", msg); goto exit_main; } } while(0)
#define TRY_SQLITE2(msg, ret_err) \
    do { if (ret_err != SQLITE_OK) { printf("error %s\n", msg);}} while(0)

// macro start/stop timer
#define START_TIMER() clock_gettime(CLOCK_MONOTONIC, &prev_tv)
#define STOP_TIMER() clock_gettime(CLOCK_MONOTONIC, &tv)
#define f_RESULT_TIME (tv.tv_sec - prev_tv.tv_sec + \
    (tv.tv_nsec - prev_tv.tv_nsec)/1.0e9)
    
//------------------------------------------------------------------------------
// Type declear
//------------------------------------------------------------------------------
typedef struct _measurement_t {
    int tag_id;
    char name[100];
    char unit[30];
    int64_t resolution;
    double ref_min;
    double ref_max;
    int64_t sampling_rate;
    char *descriptor;
    BOOL active;
} measurement_t;
//------------------------------------------------------------------------------
// Global variables
//------------------------------------------------------------------------------
// Libusb
static libusb_device_handle * LIBUSB_CALL dev_handle = NULL;
static int EP_SIZE = 8;
static int FIRMWARE_ID = FIRMWARE_ERROR;
uint8_t buffer[1024];

// Benchmark
static struct timespec tv, prev_tv;
static time_t secTime = 0;
static int counter_usb = 0;
static int counter_sqlite = 0;

// sqlite3
sqlite3 *conn;
sqlite3_stmt *stmt;
static char patientID[100];
static char db_filename[100];
static char sqlite_path[100];
static measurement_t mnt;
static struct timespec time1;
static BOOL createNewDBFile = 0;

// management and pthread
static volatile int do_exit = 0;
static int counter_wal = 0;
pthread_mutex_t lock1_mutex;
pthread_cond_t wal_checkout_cv;


//------------------------------------------------------------------------------
// function prototype
//------------------------------------------------------------------------------
// Libusb
int get_firmware_id(uint8_t, uint8_t);
static void scan_devices();
static int init_xfer(uint8_t);
static void LIBUSB_CALL callback_transfer(struct libusb_transfer *);

// management
int is_file_exist(char*); 
int parse_opt(int, char**);
void signal_handler (int);
void clean_up();

// Sqlite3
void init_sqlite();
void cache_prepare_sqlite();
void finalize_sqlite();
void *commit_thread_main(void*); 
//------------------------------------------------------------------------------
// Main function
//------------------------------------------------------------------------------
int main(int argc, char **argv)
{
    int ret_err = 0;
    pthread_t commit_thread;
    pthread_attr_t attr;
    void *status;
    
    // register signal handler
    signal(SIGTERM, signal_handler);
    signal(SIGINT, signal_handler);
    signal(SIGBREAK, signal_handler);
    #ifdef SIGKILL
    signal(SIGKILL, signal_handler);
    #endif
    #ifdef SIGABRT
    signal(SIGABRT, signal_handler);
    #endif
    
    // parse argument
    if (parse_opt(argc, argv)) exit(1);
    printf("[Patient ID %s]\n", patientID);
    
    // init usb context
    TRY_LIBUSB("initial usb context", libusb_init(NULL));
    
    //debug all usb message
	#ifdef DEBUG
	libusb_set_debug(NULL, 4);
	printf("Enable Debug level 4\n");
	#endif
        
    scan_devices();
    if (dev_handle == NULL) {printf("No device"); goto exit_main;}
    
    
     // Delete old test.db
	if (createNewDBFile && is_file_exist(db_filename))
	{
		printf("Delete previous %s\n", db_filename);
		if (remove(db_filename) != 0) {
			printf("Cannot remove file\n");
			exit(1);
		}
	}
    
    // initialize sqlite context
    init_sqlite();
    cache_prepare_sqlite();
    
    // initialize mutex and condition variable 
    pthread_mutex_init(&lock1_mutex, NULL);
    pthread_cond_init(&wal_checkout_cv, NULL);
    
    // initialize commit_thread
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);
    TRY_STOP("initial sqlite_thread", pthread_create(
        &commit_thread, &attr, commit_thread_main, NULL));
    pthread_attr_destroy(&attr);
    
    printf("[IO start]\n");
    TRY_LIBUSB("claming interface 0", libusb_claim_interface(dev_handle, 0));
    START_TIMER();
    init_xfer(EP_ADDRESS);
    while(!do_exit)
    {
        TRY_LIBUSB2("handle event", libusb_handle_events(NULL));
    }
    
    // wait for commit thread
    TRY("wait for sqlite_thread exit", pthread_join(commit_thread, &status));
exit_main:        
    #ifndef __MINGW32__
    #ifndef __MINGW64__
    clean_up();
    finalize_sqlite();
    #endif
    #endif
    
    libusb_exit(NULL);
    pthread_exit(NULL);
    return ret_err;
}

// function to check the exist file
// 0 is not; 1 is exists
int is_file_exist(char *filename) 
{
    FILE *file;
    if ((file = fopen(filename, "r")) == NULL)
    {
        if (errno == ENOENT) {
            printf("File %s doesn't exist\n", filename);
        } else {
            printf("Some error occurs with fopen(%s)\n", filename);
        }
    } else {
        fclose(file);
        return 1;
    }
    return 0;
}

// Signal handler: handle SIGINT from keyboard
void signal_handler (int param)
{
    printf("\nSIGNAL %d\n", param);
    do_exit = 1;
    #if defined(__MINGW32__) || defined(__MINGW64__)
    clean_up();
    finalize_sqlite();
    #endif
}

// scan_device: list usb device on host
static void scan_devices()
{
    libusb_device **devs;
    libusb_device *dev;
    int i=0, count;
    int ret_err = 0;

    count = libusb_get_device_list(NULL, &devs);
    if (count < 0) return;
    
    while ( (dev = devs[i++]) != NULL )
    {
        struct libusb_device_descriptor desc;
        int r = libusb_get_device_descriptor(dev, &desc);
        if (r < 0) {
            printf("Fail to get device descriptor\n");
            return;
        }
        
        // List all device descriptor
        /*
        printf("%04X:%04X (bus:%d, device:%d)\n",
            desc.idVendor,
            desc.idProduct,
            libusb_get_bus_number(dev),
            libusb_get_device_address(dev));
        */
        
        if (desc.idVendor == ID_VENDOR && desc.idProduct == ID_PRODUCT)
        {
            TRY_LIBUSB2("open device", libusb_open(dev, &dev_handle));
            if (ret_err)
                continue;
            int fid = get_firmware_id(desc.iManufacturer, desc.iProduct);
            if (fid == FIRMWARE_ERROR)
                continue;
            
            FIRMWARE_ID = fid;
            EP_SIZE = libusb_get_max_packet_size(dev, EP_ADDRESS);
            printf("[IO firmware %d] [open 0x%02X (%d bytes)]\n",
                fid, EP_ADDRESS, EP_SIZE);
            
        }
    }
    libusb_free_device_list(devs, 1);
}

// Get firmware Id from string descriptor
int get_firmware_id(uint8_t im, uint8_t ip)
{
    // default fid indicating no firmware
    int fid = FIRMWARE_ERROR;
    int ret_err = 0;
    unsigned char *manufacturer = (unsigned char*)malloc(sizeof(unsigned char)*201);
    unsigned char *product = (unsigned char*)malloc(sizeof(unsigned char)*201);
    
    if (dev_handle == NULL) goto exit_firmware;

    ret_err = libusb_get_string_descriptor_ascii(
        dev_handle, im, manufacturer, 201);
    if (ret_err < 0) {
        printf("err libusb %d:%s", ret_err, libusb_error_name(ret_err));
        goto exit_firmware;
    }
    
    ret_err = libusb_get_string_descriptor_ascii(
        dev_handle, ip, product, 201);
    if (ret_err < 0) {
        printf("err libusb %d:%s", ret_err, libusb_error_name(ret_err));
        goto exit_firmware;
    }
    
    printf("[Device %s (%s)]\n", manufacturer, product);
    if (strncasecmp(manufacturer, FIRMWARE_CA_TEST__MF, FIRMWARE_CA_TEST__MF_LEN) == 0
        && strncasecmp(product, FIRMWARE_CA_TEST__PD, FIRMWARE_CA_TEST__PD_LEN) == 0)
    {
        fid = FIRMWARE_CA_TEST;
        mnt.ref_min = 0;
        mnt.ref_max = 100;
        mnt.resolution = 100;
        mnt.descriptor = (char*)malloc(sizeof(char)*strlen("{\"id\", \"temperature\"}")+10);
        sprintf(mnt.name, "general");
        sprintf(mnt.unit, "Celcius");
        sprintf(mnt.descriptor, "{\"id\", \"temperature\"}");
    }
    if (strncasecmp(manufacturer, FIRMWARE_CA_ECG_MONITOR__MF, FIRMWARE_CA_ECG_MONITOR__MF_LEN) == 0
        && strncasecmp(product, FIRMWARE_CA_ECG_MONITOR__PD, FIRMWARE_CA_ECG_MONITOR__PD_LEN) == 0)
    {
        fid = FIRMWARE_CA_ECG_MONITOR;
        mnt.ref_min = 0;
        mnt.ref_max = 3.3;
        mnt.resolution = 16777216;
        mnt.descriptor = (char*)malloc(sizeof(char)*strlen(
            "{\"Lead-I\", \"Lead-II\", \"Lead-III\", \"aVR\", \"aVL\", \"aVF\",\
            \"V1\", \"V2\", \"V3\", \"V4\", \"V5\", \"V6\"}")+10);
        sprintf(mnt.name, "ecg");
        sprintf(mnt.unit, "mV");
        sprintf(mnt.descriptor, "{\"Lead-I\", \"Lead-II\", \"Lead-III\", \
            \"aVR\", \"aVL\", \"aVF\", \"V1\", \"V2\", \"V3\", \"V4\", \
            \"V5\", \"V6\"}");
    }
    if (strncasecmp(manufacturer, FIRMWARE_CA_PULSE_OXIMETER__MF, FIRMWARE_CA_PULSE_OXIMETER__MF_LEN) == 0
        && strncasecmp(product, FIRMWARE_CA_PULSE_OXIMETER__PD, FIRMWARE_CA_PULSE_OXIMETER__PD_LEN) == 0)
    {
        fid = FIRMWARE_CA_PULSE_OXIMETER;
        mnt.ref_min = 0;
        mnt.ref_max = 3.3;
        mnt.resolution = 4194304;
        mnt.descriptor = (char*)malloc(sizeof(char)*strlen("{\"LED2\", \"LED1\"}")+10);
        sprintf(mnt.name, "oxigen_sat");
        sprintf(mnt.unit, "mV");
        sprintf(mnt.descriptor, "{\"LED2\", \"LED1\"}");
    }
    
exit_firmware:
    free(manufacturer);
    free(product);
    return fid;
}

// initial interrupt xfer
static int init_xfer(uint8_t endpoint)
{
	static struct libusb_transfer *xfr;
	xfr = libusb_alloc_transfer(1);
	if (!xfr)
		return LIBUSB_ERROR_NO_MEM;
	memset(buffer, 0, sizeof(buffer));
	libusb_fill_interrupt_transfer(
		xfr, //transfer
		dev_handle, //handle
		endpoint, //target endpoint
		buffer, //buffer
		EP_SIZE, //size of buffer
		callback_transfer, //pointer callback function
		NULL, //no user data 
		0); //unlimit timeout
	return libusb_submit_transfer(xfr);
}

// callback function when URB receive data
static void LIBUSB_CALL callback_transfer(struct libusb_transfer *xfr)
{
	uint16_t i = 0;
    int ret_err;
	uint8_t *packet;
    
	// checking transfer status
	if (xfr->status != LIBUSB_TRANSFER_COMPLETED) 
	{
		printf("Error transfer status: %d\n", xfr->status);
		libusb_free_transfer(xfr);
		exit(3);
	}
    // increase packet counter
	counter_usb++;
    packet = xfr->buffer;
    
    // parse packet depended on firmware id
    switch (FIRMWARE_ID) {
        case FIRMWARE_CA_TEST:
        
        if (xfr->actual_length >= 6) {
            
            // Step 1: parse data
            int64_t channel_value[2];
            channel_value[0] = (int64_t)packet[0];
            channel_value[1] = (int64_t)packet[1];
            clock_gettime(CLOCK_REALTIME, &time1);
            
            // Step 2: sqlite insert
            for (i=0; i<2; i++)
            {
                sqlite3_reset(stmt);
                sqlite3_bind_int64(stmt, 1, (int64_t)(time1.tv_sec*1e9 + time1.tv_nsec));
                sqlite3_bind_int64(stmt, 2, i);
                sqlite3_bind_int64(stmt, 3, mnt.tag_id);
                sqlite3_bind_int64(stmt, 4, channel_value[i]);
                ret_err = sqlite3_step(stmt);
                if (ret_err != SQLITE_DONE) goto exit_xfer;
            }
            counter_sqlite++;
            
            // Step 3: wal_checkout
            pthread_mutex_lock(&lock1_mutex);
            counter_wal+=2;
            if (counter_wal >= WAL_CHECKOUT_LENGTH)
                pthread_cond_signal(&wal_checkout_cv);
            pthread_mutex_unlock(&lock1_mutex);
            
            counter_sqlite++;
        }
        break;
        case FIRMWARE_CA_PULSE_OXIMETER:
        if (xfr->actual_length >= 6) {
            
            // Step 1: parse data
            int64_t channel_value[2];
            channel_value[0] = (packet[0] << 16) & 0xFF0000;
            channel_value[0] += (packet[1] << 8) & 0x00FF00;
            channel_value[0] += (packet[2] & 0x0000FF);
            channel_value[1] = (packet[3] << 16) & 0xFF0000;
            channel_value[1] += (packet[4] << 8) & 0x00FF00;
            channel_value[1] += (packet[5] & 0x0000FF);
            clock_gettime(CLOCK_REALTIME, &time1);
            
            // Step 2: sqlite insert
            for (i=0; i<2; i++)
            {
                sqlite3_reset(stmt);
                sqlite3_bind_int64(stmt, 1, (int64_t)(time1.tv_sec*1e9 + time1.tv_nsec));
                sqlite3_bind_int64(stmt, 2, i);
                sqlite3_bind_int64(stmt, 3, mnt.tag_id);
                sqlite3_bind_int64(stmt, 4, channel_value[i]);
                ret_err = sqlite3_step(stmt);
                if (ret_err != SQLITE_DONE) goto exit_xfer;
            }
            counter_sqlite++;
            
            // Step 3: wal_checkout
            pthread_mutex_lock(&lock1_mutex);
            counter_wal+=2;
            if (counter_wal >= WAL_CHECKOUT_LENGTH)
                pthread_cond_signal(&wal_checkout_cv);
            pthread_mutex_unlock(&lock1_mutex);
        }
        break;
        case FIRMWARE_CA_ECG_MONITOR:
        if (xfr->actual_length >= 27) {
            
            // Step 1: parse data
            /* CH1 = V6
             * CH2 = Lead-I 
             * CH3 = Lead-II
             * CH4 = V2
             * CH5 = V3
             * CH6 = V4
             * CH7 = V5
             * CH8 = V1
             * Lead-III = Lead II - Lead I
             * aVR = -(Lead II + Lead I) /2
             * aVL = LeadI - LeadII/2
             * aVF = LeadII - LeadI/2
             */
            int64_t channel_value[12];
            // Lead I
            channel_value[0] = (packet[6] << 16) & 0xFF0000;
            channel_value[0] += (packet[7] << 8) & 0x00FF00;
            channel_value[0] += (packet[8] & 0x0000FF);
            // Lead II
            channel_value[1] = (packet[9] << 16) & 0xFF0000;
            channel_value[1] += (packet[10] << 8) & 0x00FF00;
            channel_value[1] += (packet[11] & 0x0000FF);
            // V1
            channel_value[6] = (packet[24] << 16) & 0xFF0000;
            channel_value[6] += (packet[25] << 8) & 0x00FF00;
            channel_value[6] += (packet[26] & 0x0000FF);
            // V2
            channel_value[7] = (packet[12] << 16) & 0xFF0000;
            channel_value[7] += (packet[13] << 8) & 0x00FF00;
            channel_value[7] += (packet[14] & 0x0000FF);
            // V3
            channel_value[8] = (packet[15] << 16) & 0xFF0000;
            channel_value[8] += (packet[16] << 8) & 0x00FF00;
            channel_value[8] += (packet[17] & 0x0000FF);
            // V4
            channel_value[9] = (packet[18] << 16) & 0xFF0000;
            channel_value[9] += (packet[19] << 8) & 0x00FF00;
            channel_value[9] += (packet[20] & 0x0000FF);
            // V5
            channel_value[10] = (packet[21] << 16) & 0xFF0000;
            channel_value[10] += (packet[22] << 8) & 0x00FF00;
            channel_value[10] += (packet[23] & 0x0000FF);
            // V6
            channel_value[11] = (packet[3] << 16) & 0xFF0000;
            channel_value[11] += (packet[4] << 8) & 0x00FF00;
            channel_value[11] += (packet[5] & 0x0000FF);
            // Lead III
            channel_value[2] = channel_value[1] - channel_value[0];
            // aVR
            channel_value[3] = -(channel_value[1]+channel_value[0])/2;
            // aVL
            channel_value[4] = channel_value[0] - channel_value[1]/2;
            // aVF
            channel_value[5] = channel_value[1] - channel_value[0]/2;
            clock_gettime(CLOCK_REALTIME, &time1);
            
            // Step 2: sqlite insert
            for (i=0; i<12; i++)
            {
                sqlite3_reset(stmt);
                sqlite3_bind_int64(stmt, 1, (int64_t)(time1.tv_sec*1e9 + time1.tv_nsec));
                sqlite3_bind_int64(stmt, 2, i);
                sqlite3_bind_int64(stmt, 3, mnt.tag_id);
                sqlite3_bind_int64(stmt, 4, channel_value[i]);
                ret_err = sqlite3_step(stmt);
                if (ret_err != SQLITE_DONE) goto exit_xfer;
            }
            counter_sqlite++;
            // Step 3: wal_checkout
            pthread_mutex_lock(&lock1_mutex);
            counter_wal+=12;
            if (counter_wal >= WAL_CHECKOUT_LENGTH)
                pthread_cond_signal(&wal_checkout_cv);
            pthread_mutex_unlock(&lock1_mutex);
        }
        break;
    }
exit_xfer:
    // re submit interrupt transfer
	if (libusb_submit_transfer(xfr) < 0)
	{
		printf("Error re-submmit transfer\n");
		exit(1);
	}
}

int parse_opt(int argc, char **argv)
{
    int index, c;
    size_t len_patient;
    while ((c = getopt(argc, argv, "n")) != -1) 
    {
        switch (c)
        {
            case 'n':
            createNewDBFile = 1;
            break;
            default:
            return 1;
        }
    }
    
    // parse patient ID
    if (optind < argc) 
    {
        len_patient = strlen(argv[optind]);
        if (len_patient > 0) {
            strncpy_s(patientID, sizeof(patientID), argv[optind], len_patient);
            sprintf(db_filename, "%s.db", patientID);
            sprintf(sqlite_path, "file:%s.db?mode=rwc", patientID);
            return 0;
        }
        printf("Empty patient ID\n");
    }
    printf("Missing patient ID\n");
    return 1;
}


 /*  Sqlite database for buffering the USB stream data
 *  Policy 1
 *    1. One Device can have multiple patientID
 *    2. One patienID can have a one sqlite db file
 *    3. One patient db file can have multiple Tag
 *    4. One Tag can have only one measurement
 *    5. One measurement can have multiple channels
 *    6. One channel can have multiple timestamp
 *    7. One timestamp can only have one value
 *
 *  Policy 2
 *    1. New Tag ID IF NO MATCH TagData AND Descriptor
 *    2. Measurement Table name := Measurement_TagID
 *
 */
/*   TAG TABLE |<-------------------TAGData ------------------------------->|
 *   --------------------------------------------------------------------------------------------------------------------
 *   | ID      | MNT   | UNIT | RESOLUTION | REFMAX | REFMIN | SamplingRate | ACTIVE  | Descriptor                      |
 *   --------------------------------------------------------------------------------------------------------------------
 *   | PRI_KEY | TEXT  | TEXT | INTEGER    |  REAL  | REAL   | INTEGER(nsec)| INTEGER | TEXT                            |
 *   --------------------------------------------------------------------------------------------------------------------
 *   e.g.
 *   --------------------------------------------------------------------------------------------------------------------
 *   | 1       | ECG1  |  mV  | 2048       |  1     |  -1    | 1000         | 0       | {"1": "LEAD_I", "2": "LEAD_II"} |
 *   --------------------------------------------------------------------------------------------------------------------
 *   | 2       | ECG2  |   V  | 1024       |  5     |   0    | 2000         | 1       | {"1": "LEAD_VI", "4": "LEAD_II"}|
 *   --------------------------------------------------------------------------------------------------------------------
 *   | 3       | SPO2  |  %   | 1024       |  0     |  100   | 3000         | 0       | {"5": "LEAD_I", "10": "LEAD_II"}|
 *   --------------------------------------------------------------------------------------------------------------------
 *  MNT : Measurement
 *  CHN : Channel
 *  TABLE_NAMES : ECG1_1, ECG2_2, SPO2_3
 */

/*   Structure MEASUREMENT Table
 *
 *   Table Name (see Policy 2.2) e.g. ECG_1
 *   --------------------------------------------
 *   | TIME    | CHANNEL_ID | VALUE   | TAG_ID  |
 *   --------------------------------------------
 *   | PRI_KEY | INTEGER    | INTEGER | INTEGER |
 *   --------------------------------------------
 *   e.g.
 *   --------------------------------------------
 *   | 109880980980  | 1    |  10908  | 1       |
 *   --------------------------------------------
 *   | 1988-09-0909  | 2    |  78909  | 1       |
 *   --------------------------------------------
 */
void init_sqlite()
{
    int ret_err;
    char *measurementTable;
    printf("[Sqlite open %s]\n", sqlite_path);
    TRY_SQLITE("open db_file failed", sqlite3_open_v2(sqlite_path, &conn,
        SQLITE_OPEN_READWRITE | SQLITE_OPEN_CREATE | SQLITE_OPEN_URI, NULL));
    TRY_SQLITE("PRAGMA optimization",
        sqlite3_exec(conn, "PRAGMA journal_mode=WAL; \
        PRAGMA wal_autocheckpoint=1000; PRAGMA auto_vacuum=FULL;"
        , NULL, NULL, NULL));
    
    /* Create table tag
	   TAG TABLE |<-------------------TAGData ------------------------------->|
	   --------------------------------------------------------------------------------------------------------------------
	   | ID      | MNT   | UNIT | RESOLUTION | REFMAX | REFMIN | SamplingRate | ACTIVE  | Descriptor                      |
	   --------------------------------------------------------------------------------------------------------------------
	   | PRI_KEY | TEXT  | TEXT | INTEGER    |  REAL  | REAL   | INTEGER(nsec)| INTEGER | TEXT                            |
	   --------------------------------------------------------------------------------------------------------------------
	   e.g.
	   --------------------------------------------------------------------------------------------------------------------
	   | 1       | ECG1  |  mV  | 2048       |  1     |  -1    | 1000         | 0       | {"1": "LEAD_I", "2": "LEAD_II"} |
	   --------------------------------------------------------------------------------------------------------------------
	   | 2       | ECG2  |   V  | 1024       |  5     |   0    | 2000         | 1       | {"1": "LEAD_VI", "4": "LEAD_II"}|
	   --------------------------------------------------------------------------------------------------------------------
	   | 3       | SPO2  |  %   | 1024       |  0     |  100   | 3000         | 0       | {"5": "LEAD_I", "10": "LEAD_II"}|
	   --------------------------------------------------------------------------------------------------------------------
	   MNT : Measurement
	   CHN : Channel
	*/
	TRY_SQLITE("create tag table", sqlite3_exec(conn, 
        "CREATE TABLE IF NOT EXISTS tag (\
        id INTEGER NOT NULL PRIMARY KEY,\
        mnt TEXT NOT NULL,\
        unit TEXT,\
        resolution INTEGER,\
        ref_min INTEGER,\
        ref_max INTEGER,\
        sampling_rate INTEGER,\
        descriptor TEXT NOT NULL,\
        active INTEGER DEFAULT 0);", NULL, NULL, NULL));
        
    // insert tag record
    TRY_SQLITE("prepare query tag", sqlite3_prepare_v2(conn, 
        "SELECT id FROM tag \
        WHERE mnt = ? AND unit = ? AND resolution = ? AND ref_min = ? \
        AND ref_max = ? AND sampling_rate = ? AND descriptor= ?;"
        , -1, &stmt, NULL));
    sqlite3_bind_text(stmt, 1, mnt.name, strlen(mnt.name), 0);
    sqlite3_bind_text(stmt, 2, mnt.unit, strlen(mnt.unit), 0);
    sqlite3_bind_int64(stmt, 3, mnt.resolution);
    sqlite3_bind_double(stmt, 4, mnt.ref_min);
    sqlite3_bind_double(stmt, 5, mnt.ref_max);
    sqlite3_bind_int64(stmt, 6, mnt.sampling_rate);
    sqlite3_bind_text(stmt, 7, mnt.descriptor, strlen(mnt.descriptor), 0);
    ret_err = sqlite3_step(stmt);
    if (ret_err == SQLITE_ROW) { // if old tag found; active the record
        mnt.tag_id = sqlite3_column_int(stmt, 0);
        
        TRY_SQLITE("prepare update tag", sqlite3_prepare_v2(conn, 
        "UPDATE tag SET active = 1 WHERE id = ?;", -1, &stmt, NULL));
        
        sqlite3_bind_int(stmt, 1, mnt.tag_id);
        
        ret_err = sqlite3_step(stmt);
        if (ret_err != SQLITE_DONE) {
            printf("Err update measurement active value\n");
            goto exit_main;
        }
        
        sqlite3_finalize(stmt);
    } else { // if tag not found; create new 
        sqlite3_finalize(stmt);
        TRY_SQLITE("prepare insert tag", sqlite3_prepare_v2(conn, "INSERT INTO tag\
        (mnt, unit, resolution, ref_min, ref_max, sampling_rate, descriptor,\
        active) VALUES (?,?,?,?,?,?,?,?);", -1, &stmt, NULL));
        sqlite3_bind_text(stmt, 1, mnt.name, strlen(mnt.name), 0);
        sqlite3_bind_text(stmt, 2, mnt.unit, strlen(mnt.unit), 0);
        sqlite3_bind_int64(stmt, 3, mnt.resolution);
        sqlite3_bind_double(stmt, 4, mnt.ref_min);
        sqlite3_bind_double(stmt, 5, mnt.ref_max);
        sqlite3_bind_int64(stmt, 6, mnt.sampling_rate);
        sqlite3_bind_text(stmt, 7, mnt.descriptor, strlen(mnt.descriptor), 0);
        sqlite3_bind_int(stmt, 8, 1);
        ret_err = sqlite3_step(stmt);
        if (ret_err == SQLITE_DONE) {
            mnt.tag_id = (int) sqlite3_last_insert_rowid(conn);
        }
        sqlite3_finalize(stmt);
    }
    
    // insert measurement table
    // Create MEASUREMENT Table
	/*   Structure MEASUREMENT Table
	 *   --------------------------------------------
	 *   | TIME    | CHANNEL_ID | VALUE   | TAG_ID  |
	 *   --------------------------------------------
	 *   | PRI_KEY | INTEGER    | INTEGER | INTEGER |
	 *   --------------------------------------------
	 *   e.g.
	 *   --------------------------------------------
	 *   | 109880980980  | 1    |  10908  | 1       |
	 *   --------------------------------------------
	 *   | 1988-09-0909  | 2    |  78909  | 1       |
	 *   --------------------------------------------
	 */
    measurementTable = (char*) malloc(300);
    sprintf(measurementTable, "CREATE TABLE IF NOT EXISTS %s_%d (\
        time INTEGER NOT NULL,\
        channel_id INTEGER NOT NULL,\
        tag_id INTEGER NOT NULL,\
        value INTEGER NOT NULL,\
        PRIMARY KEY (time, channel_id, tag_id));", mnt.name, mnt.tag_id);
	TRY_SQLITE("create measurement table", sqlite3_exec(conn, measurementTable
        , NULL, NULL, NULL));
    
exit_main:
    free(measurementTable);
    return;
}

void cache_prepare_sqlite()
{
    char *insertStmt;
    if (conn) {
        
        // initial SQL raw string
        insertStmt = (char*) malloc(sizeof(char)*500);
        sprintf(insertStmt, "INSERT INTO %s_%d \
        (time, channel_id, tag_id, value) VALUES (?,?,?,?);"
        , mnt.name, mnt.tag_id);
        
        // prepares statement
        TRY_SQLITE("cache insertion prepare stmt", 
            sqlite3_prepare_v2(conn, insertStmt, -1, &stmt, NULL));
        
        TRY_SQLITE("begin TX from (inside cache_prepare_sqlite)",
            sqlite3_exec(conn, "BEGIN;", NULL, NULL, NULL));
    }
exit_main:
    // free raw SQL
    free(insertStmt);
    return;
}

// finalize_sqlite
//  1. deselect the current measurement 
//  2. Exec last commit to database
void finalize_sqlite()
{
    int ret_err;
    sqlite3_finalize(stmt);
    TRY_SQLITE2("Disable current measurement", sqlite3_prepare_v2(conn, 
        "UPDATE tag SET active = 0 WHERE id = ?;", -1, &stmt, NULL));
    sqlite3_bind_int(stmt, 1, mnt.tag_id);
    sqlite3_step(stmt);
    sqlite3_finalize(stmt);
    TRY_SQLITE2("Last commit", sqlite3_exec(conn, "COMMIT;", NULL, NULL, NULL));
    sqlite3_close_v2(conn);
}
void *commit_thread_main(void* t) 
{
    int ret_err;
    
    printf("[Commit-Thread start]\n");
    
    // main routine
    while(!do_exit) 
    {
        pthread_mutex_lock(&lock1_mutex);
        // wait for counter_wal > 1000
        while(counter_wal < 1000) {
            pthread_cond_wait(&wal_checkout_cv, &lock1_mutex);
            TRY_SQLITE2("COMMIT from thread", sqlite3_exec(conn, 
                "COMMIT;BEGIN;", NULL, NULL, NULL));
            counter_wal = 0;
        }
        pthread_mutex_unlock(&lock1_mutex);
    }
    
exit_main:
    printf("[Commit-Thread stop]\n");
    pthread_exit(NULL);
}

// clean up before exit
void clean_up()
{
    STOP_TIMER();
    printf("[IO Stop] [Running Time %.6f sec] [Read %d Loss %d Uncommit %d]\n", 
        f_RESULT_TIME, counter_usb, counter_usb - counter_sqlite, counter_wal);
    TRY_LIBUSB2("release interface 0", libusb_release_interface(dev_handle, 0));
    pthread_mutex_destroy(&lock1_mutex);
    pthread_cond_destroy(&wal_checkout_cv);
    free(mnt.descriptor);
    if (dev_handle != NULL) libusb_close(dev_handle);
}