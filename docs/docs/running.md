# Running

To run Mantis, you can either define a Dockerfile and use `ghcr.io/dubonzi/mantis:latest` as your base image or download the executable from `https://github.com/dubonzi/mantis/releases`.

Mantis works by reading `Mapping` definitions which are JSON files containing information about the request you want mock such as HTTP Method, URL and other attributes, and also the corresponding response for that request. 

The default base paths Mantis reads mappings and responses files from is `files/mappings` and `files/responses` respectively. You can freely add subfolders and also configure these base paths. If running on Docker, don't forget to copy your definition files into the image when building.

Check [configuration](config.md) for options. A repository with a full example can be found [here](https://github.com/dubonzi/mantis-example).