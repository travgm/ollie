Ollie Editor
============

![ollie logo](https://i.imgur.com/clAGlbL.png)

My dogs name is Ollie.

This is his editor... kind of.

A small editor that takes after the unix editor ed.

Usage
=====

You can start ollie with or without a filename
```
ollie test.txt
```
and then it will drop you right into the editor. You can then begin typing whatever you like, each line is added to a buffer after you press enter each time. To exit editing text and go to the command prompt type ```.``` on its own line and press enter.

To save the file just type ```w``` if you started ollie with a filename. If you want to write everything to a different file
you can type ```w test.txt``` and it will save to that file instead.

To exit simply type ```q``` at the command prompt.
