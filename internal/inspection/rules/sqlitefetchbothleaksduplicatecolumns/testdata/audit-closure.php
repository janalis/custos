<?php
function f(SQLite3Result $r){$row=$r->fetchArray();echo json_encode($row);$f=function()use($row){echo $row[0];};$f();}
