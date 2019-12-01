# YAHAS
Yet Another Home Automation Server is a lightweight home automation server written in GO.
Yahas makes use of plugins providing great modularity, allowing it to be easily extended.

For more detailed info consult the [wiki](https://github.com/zechenturm/yahas/wiki)

# OS Compatibility

Yahas currently run on Linux and maxOS (untested). This is a limitation of the [GO plugin package](https://golang.org/pkg/plugin/).

## Run

You will need a recent installation of go with module support. Go 1.11+ should work. You also need `make`.

Get yahas with git:

    git clone https://github.com/zechenturm/yahas.git
    
Now build it with

    make
  
or build and run directly using

    make run
