# libs5-go

This library was originally implemented for the v0 spec of [S5](https://s5.pro/). It has a lot of warts and some WIP refactors in the `refactor` branch given this was created when I treated Go more like traditional OOP and it didn't yet click that interfaces are inferred vs explict like PHP, TS, Java. I have gotten a lot better at go since 😅.

Based on the stated direction of the S5 project per https://forum.sia.tech/t/s5-v1-rewrite-it-in-rust-large-grant-proposal/947/1, this repo is being archived as any future support in Lume will benefit from a repo that starts from 0 and ports over the application-level protocol implementation from what will be the Rust/TS versions.

The [portal-plugin-s5](https://github.com/LumeWeb/portal-plugin-s5) will stay as is as that is largely the portal side support, and not the protocol side support, and will be updated when work continues on S5.

As this is MIT, feel free to take what you want from this code base.

Happy hacking/buidling :)
