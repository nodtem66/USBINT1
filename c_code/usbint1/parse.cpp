#include "cpptoml.h"

#include <iostream>
#include <cassert>

using namespace cpptoml;

int main(int argc, char** argv)
{
    if (argc < 2)
    {
        std::cout << "Usage: " << argv[0] << " filename" << std::endl;
        return 1;
    }

    try
    {
        table g = parse_file(argv[1]);
        std::shared_ptr<base> b = g.get_qualified("server.address");
        //std::cout <<  << std::endl;
    }
    catch (const parse_exception& e)
    {
        std::cerr << "Failed to parse " << argv[1] << ": " << e.what() << std::endl;
        return 1;
    }

    return 0;
}
