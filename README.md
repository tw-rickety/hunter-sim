# (WIP) TurtleWoW Marksman Hunter DPS Simulator

This is a basic API that accepts a hunter's stats (AP, Crit, etc) and returns simulated DPS output. 

## How do I use it?
This is not currently deployed, you have to run it locally to test.

### To run locally (Windows):
* I recommend installing [Windows Subsystem For Linux](https://learn.microsoft.com/en-us/windows/wsl/install) by running this in your terminal:
    ```
    wsl --install
    ```
then open the WSL terminal and follow the steps below

### To run locally (linux/mac os/windows WSL):
* Clone or download this repository

* Install Go 1.21 or later:
  * **Linux**: Download from [golang.org/dl](https://golang.org/dl/) or use your package manager:
    ```bash
    # Ubuntu/Debian/Windows WSL
    sudo apt update
    sudo apt install golang-go
    ```
  * Verify installation:
    ```bash
    go version
    ```


* Run the simulator in a terminal window (make sure to `cd` into the `hunter-sim` folder first):
  ```bash
  make run
  ```
* Open a new terminal window (WSL terminal if you're on Windows) and run the sample request in [example_curl.txt](example_curl.txt). Modify values to whatever you want (I took these from my in-game character panel)
    ```
    curl -X POST http://localhost:8080/simulate   -H "Content-Type: application/json"   -d '{
        "AP": 1751,
        "Crit": 31.48,
        "Hit": 7,
        "ItemHaste": 1.0,
        "QuiverHaste": 1.15,
        "ArrowDPS": 20,
        "Bow": {
        "MinDamage": 144,
        "MaxDamage": 255,
        "ScopeDamage": 7,
        "Speed": 3.1
        },
    "Talents": {
        "RangedWeaponSpec": 1.05,
        "SwiftReflexesHaste": 1.02
        },
        "Race": {
        "Haste": 1.01
        },
        "MultishotCooldown": 9,
        "Ping": 150,
        "NumberOfSims": 10000,
        "FightDurationInSeconds": 60
    }' | json_pp
    ```



## How does it work?
The simulator tries to emulate game logic as closely as possible. It steps through a game loop every 10ms (`clock.go`) and follows a typical hunter autoshot, steadyshot, multishot,  weaving rotation, queued by `actionqueue.go`. Randomized damage, based on real 1.12.1 ranged attack formulas, is calculated in `hunter.go`.

To check accuracy, three things were done:
* I compared calculated autoshot damage range and attack speed to damage range in in-game character panel (check out `hunter_test.go`)

![Char Panel Damage Range](./images/char-panel.png)
![Char Panel Compare](./images/char-panel-compare.png)

* I compared calculated multishot and steadyshot damage ranges to actual damage from hitting the level 1 dummy for very long time
* I compared general damage share from each ability to actual raid logs to make sure skill damage percentages were similar (keep in mind, the sim doesn't have pet damage, so numbers will be slightly different with similar ratios):

![turtlogs](./images/turtlogs.png)
![turtlogs comparison](./images/turtlogs-compare.png)



## TODO list:
* build frontend to select gear, calculate base stats, send sim request to API, and display results
* implement trinket swapping logic and passing in specials such as trinkets
* implement armor (currently, armor/mob level is not taken into account)
* implement pet DPS simulation

