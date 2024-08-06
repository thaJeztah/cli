# docker save

<!---MARKER_GEN_START-->
Save one or more images to a tar archive (streamed to STDOUT by default)

### Aliases

`docker image save`, `docker save`

### Options

| Name             | Type     | Default | Description                                                                                                                                                                                                                                                      |
|:-----------------|:---------|:--------|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-o`, `--output` | `string` |         | Write to a file, instead of STDOUT                                                                                                                                                                                                                               |
| `--platform`     | `string` |         | Specify a platform from a multi-platform image to save.<br>If a platform is not specified, and the image is a multi-platform image, all platform variants are saved.<br><br>Format: `os[/arch[/variant]]`<br>Example: `docker image save --platform linux/amd64` |


<!---MARKER_GEN_END-->

