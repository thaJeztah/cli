# image inspect

<!---MARKER_GEN_START-->
Display detailed information on one or more images

### Options

| Name             | Type     | Default | Description                                                                                                                                                                                                                                                        |
|:-----------------|:---------|:--------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-f`, `--format` | `string` |         | Format output using a custom template:<br>'json':             Print in JSON format<br>'TEMPLATE':         Print output using the given Go template.<br>Refer to https://docs.docker.com/go/formatting/ for more information about formatting output with templates |
| `--platform`     | `string` |         | Inspect a specific platform of the multi-platform image.<br>If the image or the server is not multi-platform capable, the command will error out if the platform does not match.<br>'os[/arch[/variant]]': Explicit platform (eg. linux/amd64)                     |


<!---MARKER_GEN_END-->
