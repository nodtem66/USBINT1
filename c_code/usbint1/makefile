RM=rm
MKDIR=mkdir
CXX=g++
CC=gcc
LDFLAGS=-lusb-1.0 -lpthread -ldl
CXXFLAGS=-c -O2 -Wall -I./
CFLAGS=-c -O2 -Wall -I./
SRC=main.c sqlite3.c
OUTDIR=./build/
OBJ=$(addprefix $(OUTDIR), $(SRC:.c=.o))
EXEC=$(OUTDIR)usbint1
.PHONY: all clean
.SECONDARY: main-build
all: pre-build main-build

pre-build:
	$(MKDIR) -p $(OUTDIR)

post-build:
	$(RM) -f $(OBJ)

main-build: target

target: $(OBJ)
	$(CC) $(LDFLAGS) $(OBJ) -o $(EXEC)
	@$(MAKE) --no-print-directory post-build


build/%.o: %.c
	$(CC) $(CFLAGS) $< -o $@

clean:
	$(RM) -f $(OBJ) $(EXEC)
	$(RM) -rf $(OUTDIR)
