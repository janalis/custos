<?php
function f(SQLite3Result $r) { $row=$r->fetchArray(); $r->finalize(); }
