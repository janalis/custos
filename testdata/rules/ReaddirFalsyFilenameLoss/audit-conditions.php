<?php
if (<warning descr="Compare the directory entry with false.">!readdir($h)</warning>){}
if (<warning descr="Compare the directory entry with false.">readdir($h)==false</warning>){}
if (<warning descr="Compare the directory entry with false.">true==($entry=readdir($h))</warning>){}
