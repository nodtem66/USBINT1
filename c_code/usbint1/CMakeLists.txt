cmake_minimum_required(VERSION 2.8.11)

project(usbint1)



include_directories(
    D:/work/go/src/github.com/nodtem66/usbint1/c_code/usbint1/
    D:/work/go/src/github.com/nodtem66/usbint1/c_code/usbint1/

)

add_definitions(
    -DNDEBUG
)

# Linker options
set(CMAKE_LDFLAGS "${CMAKE_LDFLAGS}  -lusb-1.0 -L./lib64/")

# Library path
set(CMAKE_LDFLAGS "${CMAKE_LDFLAGS} -L\".\" ")

# Define the C sources
set ( C_SRCS
    main.c
    sqlite3.c
)

set_source_files_properties(
    ${C_SRCS} PROPERTIES COMPILE_FLAGS 
    " -O2 -Wall")

add_executable(usbint1 ${CXX_SRCS} ${C_SRCS})

