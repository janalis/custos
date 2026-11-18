<?php
function valid(PDOStatement $s){json_encode($s->fetchAll(PDO::FETCH_ASSOC));json_encode($s->fetchAll($mode));json_encode($data);strlen("json");}
function unknown($s){json_encode($s->fetchAll(PDO::FETCH_BOTH));}
function reshaped(PDOStatement $s){$rows=$s->fetchAll(PDO::FETCH_BOTH);foreach($rows as &$row){$row=$row['name'];}return json_encode($rows);}
function reshapedKeyed(PDOStatement $s){$rows=($s->fetchAll(PDO::FETCH_BOTH));foreach(($rows) as $i=>&$row){$row=$row['name'];}unset($row);return json_encode(($rows));}
function reshapedDestructured(PDOStatement $s){$rows=$s->fetchAll(PDO::FETCH_BOTH);foreach($rows as ['name'=>&$name,0=>$id]){$name=$id;}return json_encode($rows);}
