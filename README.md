# GNODE

Project Level NodeJs version manager.

.gnode
```
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

Run yarn at the same time `$ gnode yarn --version`, if node is not install in
`~/sdk`, it will install it automatically, along with the global modules.

It's currently compatible with Windows, Linux and Mac.

Mac has not been tested, yet.

## Note

This is still in prototype stage.