<?php
$p=new PDO('sqlite::memory:');$s=$p->query("SELECT 'a' AS kind, 7 AS n");$rows=$s->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);$rows['a'][0]['kind']='a';echo $rows['a'][0]['kind'];

function guarded_access($raw, $uid) {
$db=new PDO('sqlite::memory:');$s=$db->query("SELECT 'a' AS kind,7 AS n");$p=$s->fetchAll(PDO::FETCH_GROUP|PDO::FETCH_ASSOC);
isset($p['a'][0]['kind']); empty($p['a'][0]['kind']); echo $p['a'][0]['kind'] ?? 'fallback';
$p['a'][0]['kind']['nested']='value';
unset($p['a'][0]['kind']);
}
