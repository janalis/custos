<?php
$pdo=new PDO("sqlite::memory:");$s=$pdo->prepare("SELECT ?");$s->execute([1]);$s->execute($unknown);$s->fetch();
function direct(PDOStatement $s){$s->execute([[1,2]]);}
