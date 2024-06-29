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
and then it will drop you right into the editor. You can then begin typing whatever you like, each line is added to a buffer after you press enter each time. In ed it prints how many bytes are written to the file after each line, in ollie it writes the line number you're on after you type each line. To exit editing text and go to the command prompt type ```.``` on its own line and press enter.

To save the file just type ```w``` if you started ollie with a filename. If you want to write everything to a different file
you can type ```w test.txt``` and it will save to that file instead.

After every command it will drop you back to the editor to type text. If you see the line preceded with ```?``` that means you are in command mode. You can type ```a``` to go back to appending text.

To exit simply type ```q``` at the command prompt.

contributing
==========

This project was really just written for my own edification and learning, but was inspired by the simple ed text editor for unix. With that that being said all features, suggestions and patches are welcome on the ollie editor mailing list [here](https://lists.sr.ht/~travgm/ollie-editor)
