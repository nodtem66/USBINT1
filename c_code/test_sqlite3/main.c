#include <stdio.h>
#include <stdlib.h>
#include <signal.h>
#include <sys/time.h>
#include <string.h>
#include <errno.h>
#include "sqlite3.h"

#define MAXITERS 1000

// macro to handle ERROR
#define TRY(ret_err, msg) \
	do { if (ret_err != SQLITE_OK) { printf(msg); goto out; } } while(0)

// macro start/stop timer
#define START_TIMER() gettimeofday(&prev_tv, &tz);
#define STOP_TIMER() \
	gettimeofday(&tv, &tz); \
	printf("[time %.6f sec]\n", tv.tv_sec - prev_tv.tv_sec + \
	(tv.tv_usec - prev_tv.tv_usec)/1.0e6);

// global variables
static struct timeval tv, prev_tv;
static struct timezone tz;
int time_value = 0;
int channel_id = 0;
int value = 0;

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
	
int main(int argc, char **argv)
{
	sqlite3 *conn;
	char **err_msg = NULL;

	// Delete old test.db
	if (is_file_exist("test.db"))
	{
		printf("Delete previous test.db\n");
		if (remove("test.db") != 0) {
			printf("Cannot remove file\n");
			exit(1);
		}
	}
	
	// Create new test.db
	TRY(sqlite3_open_v2("file:test.db?mode=rwc", &conn, 
		SQLITE_OPEN_READWRITE | SQLITE_OPEN_CREATE | SQLITE_OPEN_URI, NULL),
		"Open test.db failed\n");
	
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
	 printf("Create SCHEMA ");
	TRY(sqlite3_exec(conn, "CREATE TABLE IF NOT EXISTS measurement ( \
		time INTEGER NOT NULL, \
		channel_id INTEGER NOT NULL, \
		value INTEGER NOT NULL, \
		PRIMARY KEY(time, channel_id));", NULL, NULL, NULL), "[FAIL]");
	printf("[OK]\nPrepare Stmt ");
	
	sqlite3_stmt* stmt;
	TRY(sqlite3_prepare_v2(conn, "INSERT INTO measurement \
		(time, channel_id, value) VALUES (?,?,?)", -1, &stmt, NULL),
		"[FAIL]");
	printf("[OK]\nInsert %d Records ", MAXITERS);
	int i;	
	START_TIMER();
	for (i=0; i<MAXITERS; i++) 
	{
		int ret;
		sqlite3_reset(stmt);
		TRY(sqlite3_bind_int64(stmt, 1, time_value++), "Bind int64[1] failed");
		TRY(sqlite3_bind_int64(stmt, 2, channel_id), "Bind int64[2] failed");
		TRY(sqlite3_bind_int64(stmt, 3, 0), "Bind int64[3] failed");
		
		if ((ret=sqlite3_step(stmt)) != SQLITE_DONE)
		{
			printf("[1] Step Error [time_value %d]", time_value);
			goto out;
		}
	}
	STOP_TIMER();
	
	printf("Insert %d With TX for group of Stmt ", MAXITERS);
	START_TIMER();
	TRY(sqlite3_exec(conn, "BEGIN;", NULL, NULL, NULL), "TX Begin failed");
	for (i=0; i<MAXITERS; i++) 
	{
		int ret;
		sqlite3_reset(stmt);
		TRY(sqlite3_bind_int64(stmt, 1, time_value++), "Bind int64[1] failed");
		TRY(sqlite3_bind_int64(stmt, 2, channel_id), "Bind int64[2] failed");
		TRY(sqlite3_bind_int64(stmt, 3, 0), "Bind int64[3] failed");
		
		if ((ret=sqlite3_step(stmt)) != SQLITE_DONE)
		{
			printf("[1] Step Error [time_value %d]", time_value);
			goto out;
		}
	}
	TRY(sqlite3_exec(conn, "COMMIT;", NULL, NULL, NULL), "TX Commit failed");
	STOP_TIMER();
	
	printf("PRAGMA journal_mode=WAL ");
	TRY(sqlite3_exec(conn, "PRAGMA journal_mode=WAL;", NULL, NULL, NULL), "[FAIL]");
	printf("[OK]\nBenchmark INSERT %d after using WAL", MAXITERS);
	
	START_TIMER();
	for (i=0; i<MAXITERS; i++) 
	{
		int ret;
		sqlite3_reset(stmt);
		TRY(sqlite3_bind_int64(stmt, 1, time_value++), "Bind int64[1] failed");
		TRY(sqlite3_bind_int64(stmt, 2, channel_id), "Bind int64[2] failed");
		TRY(sqlite3_bind_int64(stmt, 3, 0), "Bind int64[3] failed");
		
		if ((ret=sqlite3_step(stmt)) != SQLITE_DONE)
		{
			printf("[1] Step Error [time_value %d]", time_value);
			goto out;
		}
	}
	STOP_TIMER();
	
	printf("Benchmark INSERT %d after using WAL with TX at each Stmt ", MAXITERS);
	START_TIMER();
	for (i=0; i<MAXITERS; i++)
	{
		int ret;
		sqlite3_reset(stmt);
		TRY(sqlite3_bind_int64(stmt, 1, time_value++), "Bind int64[1] failed");
		TRY(sqlite3_bind_int64(stmt, 2, channel_id), "Bind int64[2] failed");
		TRY(sqlite3_bind_int64(stmt, 3, 0), "Bind int64[3] failed");
		TRY(sqlite3_exec(conn, "BEGIN;", NULL, NULL, NULL), "TX Begin failed");
		if ((ret=sqlite3_step(stmt)) != SQLITE_DONE)
		{
			printf("[1] Step Error [time_value %d]", time_value);
			goto out;
		}
		TRY(sqlite3_exec(conn, "COMMIT;", NULL, NULL, NULL), "TX Commit failed");
	}
	STOP_TIMER();
	
	printf("Benchmark INSERT %d with TX for group of 100 Stmt ", MAXITERS);
	START_TIMER();
	for (i=0; i<MAXITERS; i++)
	{
		int ret;
		sqlite3_reset(stmt);
		TRY(sqlite3_bind_int64(stmt, 1, time_value++), "Bind int64[1] failed");
		TRY(sqlite3_bind_int64(stmt, 2, channel_id), "Bind int64[2] failed");
		TRY(sqlite3_bind_int64(stmt, 3, 0), "Bind int64[3] failed");
		if ((i % 100) == 0) {
			TRY(sqlite3_exec(conn, "BEGIN;", NULL, NULL, NULL), "TX Begin failed");
		}
		if ((ret=sqlite3_step(stmt)) != SQLITE_DONE)
		{
			printf("[1] Step Error [time_value %d]", time_value);
			goto out;
		}
		if ((i % 100) == 99) {
			TRY(sqlite3_exec(conn, "COMMIT;", NULL, NULL, NULL), "TX Commit failed");
		}
	}
	
	STOP_TIMER();
	
	printf("Benchmark INSERT %d with TX for group of 1000 Stmt ", MAXITERS);
	START_TIMER();
	TRY(sqlite3_exec(conn, "BEGIN;", NULL, NULL, NULL), "TX Begin failed");
	for (i=0; i<MAXITERS; i++)
	{
		int ret;
		sqlite3_reset(stmt);
		TRY(sqlite3_bind_int64(stmt, 1, time_value++), "Bind int64[1] failed");
		TRY(sqlite3_bind_int64(stmt, 2, channel_id), "Bind int64[2] failed");
		TRY(sqlite3_bind_int64(stmt, 3, 0), "Bind int64[3] failed");
		if ((ret=sqlite3_step(stmt)) != SQLITE_DONE)
		{
			printf("[1] Step Error [time_value %d]", time_value);
			goto out;
		}
	}
	TRY(sqlite3_exec(conn, "COMMIT;", NULL, NULL, NULL), "TX Commit failed");
	STOP_TIMER();
out:
	if (stmt) {
		sqlite3_finalize(stmt);
	}
	if (err_msg) {
		sqlite3_free(err_msg);
	}
	if (conn) {
		sqlite3_close_v2(conn);	
	}
	printf("\nBye...");
	return 0;
}
