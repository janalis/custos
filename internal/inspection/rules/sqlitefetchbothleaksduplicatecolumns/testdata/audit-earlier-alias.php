<?php
function f(SQLite3Result $r){$alias=&$row;$row=$r->fetchArray();echo json_encode($row);echo $alias[0];}
