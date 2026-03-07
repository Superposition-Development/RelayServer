# Relay Server

Relay is a messaging application designed to be secure, annonymous and completely free (as in freedom!)

For now this readme will be used as placeholder documentation

# `init.py`

`init.py` is ran from `main.py` and is responsible for setting up global variables, aditionally it figures out the _system path_ used to store the database.

If `USING_CUSOM_DB_PATH` is `"False"`, the database file (default name of `d9xb1.db`) is stored at the following path:

- `nt` (Windows):
    - `%LOCALAPPDATA\relay`

- `posix` (\*nix, \*BSD)
    - `$XDG_DATA_HOME/relay`

# `main.py`

`main.py` is responsible for starting up the server itself (via `app.run(PORT)`) and defining the avaiable routes, which are:

- `/`
- `/bpAuth`
- `/bpServer`
- `/bpChannel`
- `/bpMessage`

All of these routes are implemented in their respective `.py` file inside the `routes` directory, with the exception of the `/` route, which is in `main.py`

The `/` route is defined in the `home` function, which returns a string as the `html` for the frontend.


