<?php
function f(SQLite3Stmt $s) {  $r=$s->execute(); if($r->fetchArray()){ $s->clear(); $s->bindValue(1,'later'); <warning descr="Reset the SQLite statement before rebinding cleared parameters.">$s->execute()</warning>; } }
