#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <errno.h>
#include <sys/time.h>
#include <libusb-1.0/libusb.h>

//------------------------------------------------------------------------------
// Defines 
//------------------------------------------------------------------------------
#define ID_VENDOR		0x10C4
#define ID_PRODUCT		0x8846
#define EP_INT_IN		0x83
#define EP_PACKET_SIZE	64
#define TRUE 1
#define FALSE 0
//------------------------------------------------------------------------------
// Global variables
//------------------------------------------------------------------------------
static int do_exit = FALSE;
static libusb_device_handle * LIBUSB_CALL dev_handle = NULL;
static struct timeval tv, prev_tv;
static time_t secTime = 0;
static struct timezone tz;
static int counter = 0;
uint8_t buffer[EP_PACKET_SIZE];

//------------------------------------------------------------------------------
// Signal handler: handle SIGINT from keyboard
//------------------------------------------------------------------------------
void signal_handler (int param)
{
	printf("SIGNAL %d", param);
	do_exit = TRUE;
}
//------------------------------------------------------------------------------
// scan_device: list usb device on host
//------------------------------------------------------------------------------
static void scan_devices()
{
	libusb_device **devs;
	libusb_device *dev;
	int i=0, count;
	
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
		printf("%04X:%04X (bus:%d, device:%d)\n",
			desc.idVendor,
			desc.idProduct,
			libusb_get_bus_number(dev),
			libusb_get_device_address(dev));
			
		if (desc.idVendor == ID_VENDOR && desc.idProduct == ID_PRODUCT)
		{
			printf("endpoint: %02X (%d %d)\n",
				EP_INT_IN,
				libusb_get_max_packet_size(dev, EP_INT_IN),
				libusb_get_max_iso_packet_size(dev, EP_INT_IN));
		}
	}
	libusb_free_device_list(devs, TRUE);
}

//------------------------------------------------------------------------------
// callback function when URB receive data
//------------------------------------------------------------------------------
static void LIBUSB_CALL callback_transfer(struct libusb_transfer *xfr)
{
	uint16_t i, len = 0;
	
	
	if (xfr->status != LIBUSB_TRANSFER_COMPLETED) 
	{
		printf("Error transfer status: %d\n", xfr->status);
		libusb_free_transfer(xfr);
		exit(3);
	}
	counter++;
	/*/
	printf("[%d] XFR length: %u, actual_length: %u\n",
		counter,
		xfr->length,
		xfr->actual_length);

	for (i=0, len=xfr->actual_length; i<len; i++)
	{
		printf("%02d ", xfr->buffer[i]);
	}
	printf("\n");

	gettimeofday(&tv, &tz);
	secTime = tv.tv_sec - prev_tv.tv_sec;
	if (secTime > 0)
	{
		printf("%ld ", secTime*1000000 + tv.tv_usec - prev_tv.tv_usec);
	} else {
		printf("%ld ", tv.tv_usec - prev_tv.tv_usec);
	}
	gettimeofday(&prev_tv, &tz);
	*/
	if (libusb_submit_transfer(xfr) < 0)
	{
		printf("Error re-submmit transfer\n");
		exit(1);
	}
}

//------------------------------------------------------------------------------
// test interrupt pipe 
//------------------------------------------------------------------------------
static int test_int(uint8_t endpoint)
{
	static struct libusb_transfer *xfr;
	xfr = libusb_alloc_transfer(1);
	if (!xfr)
		return LIBUSB_ERROR_NO_MEM;
	memset(buffer, 3, sizeof(buffer));
	libusb_fill_interrupt_transfer(
		xfr, //transfer
		dev_handle, //handle
		endpoint, //target endpoint
		buffer, //buffer
		sizeof(buffer), //size of buffer
		callback_transfer, //pointer callback function
		NULL, //no user data 
		0); //unlimit timeout
	return libusb_submit_transfer(xfr);
}

//------------------------------------------------------------------------------
// Main:
// run scandevice and then start to receive data from USB device
//------------------------------------------------------------------------------
#define TRY(desc, expr) err_code = expr; \
	if (err_code != 0) {\
	printf("Error %s: %s\n", desc, libusb_error_name(err_code)); goto out;} 
	
int main()
{
	int err_code = 0;
	//init signal handler
	signal(SIGINT, signal_handler);
	signal(SIGTERM, signal_handler);
	signal(SIGABRT_COMPAT, signal_handler);
	
	//init libusb
	TRY("initializing libusb", libusb_init(NULL));
	
	//debug all usb message
	#ifdef DEBUG
	libusb_set_debug(NULL, 4);
	printf("Enable Debug level 4\n");
	#endif

	//print version libusb
	const struct libusb_version *version = libusb_get_version();
	printf("libusb version: %d.%d.%d\n",
		version->major,
		version->minor,
		version->micro);
		
	//print usb device list
	scan_devices();
	
	//get usb device handle for ID_VENDOR and ID_PRODUCT
	dev_handle = libusb_open_device_with_vid_pid(NULL, ID_VENDOR, ID_PRODUCT);
	if (!dev_handle)
	{
		printf("Error finding USB device %04X:%04X\n", ID_VENDOR, ID_PRODUCT);
		goto out;
	}
	printf("found USB device %04X:%04X\n", ID_VENDOR, ID_PRODUCT);
	
	//lock interface 1
	printf("claiming interface 0\n");
	TRY("claming interface 0", libusb_claim_interface(dev_handle, 0));
	
	//start test device
	gettimeofday(&prev_tv, &tz);
	TRY("testing", test_int(EP_INT_IN));
	
	while (!do_exit)
	{
		TRY("handle event", libusb_handle_events(NULL));
	}
	gettimeofday(&tv, &tz);
	secTime = tv.tv_sec - prev_tv.tv_sec;
	if (secTime > 0)
	{
		printf("[time %.6f s]", (secTime*1000000 + tv.tv_usec - prev_tv.tv_usec)/1e6 );
	} else {
		printf("[time %.6f s]", (tv.tv_usec - prev_tv.tv_usec)/1e6);
	}
	printf(" [count %d]\n", counter);
	//release interface 1
	printf("release interface 0\n");
	libusb_release_interface(dev_handle, 0);
	
out:
	if (dev_handle)
	{
		libusb_close(dev_handle);
	}
	libusb_exit(NULL);
	return err_code;
}
