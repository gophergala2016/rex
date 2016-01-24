#REx development

Development should all be done with the locked dependency versions in
glide.lock.  To install dependencies, make sure that the `GO15VENDOREXPERIMENT`
environment variable is enabled is enabled.

    export GO15VENDOREXPERIMENT=1

Then make sure glide is up to date and use it to synchronize the `vendor/`
directory with the lock file.

    go get -u github.com/Masterminds/glide
    glide install

To run the lark.lua script(s) in the repository install the
[lark](https://github.com/bmatsuo/lark) tool.  The use the command as you would
a `make`.

    lark test
    lark ratscrew
    # ...
