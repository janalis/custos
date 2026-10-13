<?php
function f(SQLite3Result $r){$row=($r->fetchArray());echo json_encode($row);echo $row[0];}
