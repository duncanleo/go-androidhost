## go-androidhost
This is the server to Android Host. The other part of it is intellij-androidhost, which can be found [here](https://github.com/duncanleo/intellij-androidhost).

### What is Android Host?
The Android Emulator is the most versatile emulator available - you can set any screen size, and run versions of Android such as Froyo (2.2) and Marshmallow (6.0). However, it's also the slowest, even with Intel HAXM acceleration.

Android Host allows you to run the Android Emulator on a separate machine. The server runs on that separate machine and provides remote deployment and control.

I built this during my time as an intern at [buUuk](http://www.buuuk.com), where the Android developers had to test their apps on the few devices in the office. Sometimes, bugs occurred on older versions of Android of which none of the devices ran.

### Features
- Remote starting of AVDs
- Deploy to any ADB device connected to the server machine, including USB devices
- Discovery through UDP - you can deploy to multiple servers in the same LAN

## Installation
1. Clone this repo
2. `cd` to the cloned directory
3. Fetch Go dependencies

    ```shell
    go get ./...
    ```
4. Build a binary

    ```shell
    go build
    ```
5. Run the binary

    ```shell
    ./go-androidhost
    ```
6. Install the IntelliJ plugin in Android Studio. Instructions are available in the other repo.
