<?php
function f(SQLite3Stmt $s) { $s->execute(); $s->reset(); $s->clear(); $s->bindValue(1,'later'); $s->execute(); }
