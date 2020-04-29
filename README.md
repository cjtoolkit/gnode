# Gnode

Project Level NodeJs version manager.

## Installation

```
$ go get github.com/cjtoolkit/gnode
```

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

It's currently compatible with Windows, Linux and Mac.

Mac has not been tested, yet.

This is still in prototype stage.