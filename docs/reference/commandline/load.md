# docker load

<!---MARKER_GEN_START-->
Load an image from a tar archive or STDIN

### Aliases

`docker image load`, `docker load`

### Options

| Name            | Type     | Default | Description                                                                                                                                                                                                                                               |
|:----------------|:---------|:--------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-i`, `--input` | `string` |         | Read from tar archive file, instead of STDIN                                                                                                                                                                                                              |
| `--platform`    | `string` |         | Specify a platform from a multi-platform image to load.<br>If a platform is not specified, and the image is a multi-platform image, all platforms are loaded.<br><br>Format: `os[/arch[/variant]]`<br>Example: `docker image load --platform linux/amd64` |
| `-q`, `--quiet` | `bool`   |         | Suppress the load output                                                                                                                                                                                                                                  |


<!---MARKER_GEN_END-->

