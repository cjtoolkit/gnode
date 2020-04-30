# Gnode

Project Level NodeJs version manager.

## Installation

```
$ go get github.com/cjtoolkit/gnode
```

If you haven't got go installed, you can use the pre compiled binary instead  
https://github.com/cjtoolkit/gnode/releases  
Rename it to gnode or gnode.exe (if windows) and add it to the path
(e.g. `/usr/bin`).

## Using Gnode

Create `.gnode` in the root of the project, with the example below.

```json
{
  "version": "12.16.2",
  "modules": [
    {
      "package": "yarn",
      "version": "1.22.4"
    }
  ]
}
```

The first parameter of gnode will point to a file in Node's bin directory,
for example.

To use node run `$ gnode node`  
To use npm run `$ gnode npm`  
To use yarn run `$ gnode yarn`

If node is not install in `~/sdk`, it will download Node and install it
automatically, along with the modules specified in `.gnode`.

## Note

You add add `"no_npm": true` to the root of .gnode and that will strip
out npm after installation. 

It's currently compatible with Windows, Linux and Mac.

Mac has not been tested, yet.