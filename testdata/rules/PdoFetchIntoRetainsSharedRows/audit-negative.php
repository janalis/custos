<?php
$p=new PDO('sqlite::memory:'); $s=$p->query('SELECT 1 AS n UNION ALL SELECT 2'); $s->setFetchMode(PDO::FETCH_INTO,new stdClass()); $rows=[]; while($row=$s->fetch(PDO::FETCH_ASSOC)) {$rows[]=$row;} var_dump($rows);
