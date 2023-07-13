## Description

This is the Admin UI Tool created for the Ardan Labs Service Repository.

Learn more about the project:

[Wiki](https://github.com/ardanlabs/service/wiki) | [Course Outline](https://github.com/ardanlabs/service/wiki/course-outline) | [Class Schedule](https://www.eventbrite.com/o/ardan-labs-7092394651)

## To create the project we leverage the following template provided by the MUI Team

- [Material UI - Next.js App Router example in TypeScript](https://github.com/mui/material-ui/tree/41f15df97b74bbf102715857564374551fac6162/examples/material-next-app-router-ts)

## The idea behind the example

The project uses [Next.js](https://github.com/vercel/next.js), which is a framework for server-rendered React apps.
It includes `@mui/material` and its peer dependencies, including [Emotion](https://emotion.sh/docs/introduction), the default style engine in Material UI v5. If you prefer, you can [use styled-components instead](https://mui.com/material-ui/guides/interoperability/#styled-components).

### Installing the Training Material

_NOTE:_ This assumes you had clone the [Service Repository](https://github.com/ardanlabs/service).
And entered the /code folder.

```
$ make write-token-to-env
$ make admin-gui-dev
$ Navigate to localhost:3001 in your browser to see the site
```

_NOTE:_ This will install and run the server on a dev environment. For building the envrionment you have to use the following instructions.

```
$ make write-token-to-env
$ admin-gui-start-build
$ Navigate to localhost:3001 in your browser to see the site
```

_NOTE:_ This assumes you have npm and Node.JS installed. If you don’t, you can find the installation instructions here: https://docs.npmjs.com/downloading-and-installing-node-js-and-npm

_NOTE:_ Supported Operating Systems are macOS, Windows (including WSL), and Linux.
