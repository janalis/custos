<?php
function bad(PDOStatement $s){<warning descr="Use associative fetch mode before serializing rows.">json_encode($s->fetchAll(PDO::FETCH_BOTH))</warning>;}
function viaVariable(PDOStatement $s){$rows=$s->fetchAll(PDO::FETCH_BOTH);foreach($rows as $row){echo $row['name'];}foreach($other as &$o){$o=1;}foreach($rows as [$id,$name]){echo $id;}$f=function(array $rows){foreach($rows as &$row){$row=1;}return $rows;};return <warning descr="Use associative fetch mode before serializing rows.">json_encode($rows)</warning>;}
$stmt=(new PDO('sqlite::memory:'))->query('SELECT 1');
$all=$stmt->fetchAll(PDO::FETCH_BOTH);
foreach($all as &$one){$one=$one[0];}
echo json_encode($all);
