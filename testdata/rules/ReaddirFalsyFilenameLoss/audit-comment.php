<?php
if(<warning descr="Compare the directory entry with false.">!readdir(/*handle*/$h)</warning>){}
if (<warning descr="Compare the directory entry with false.">readdir($h // handle
)</warning>) {}
