<?php
// PDO accepts a named placeholder with or without its leading colon.
function reboundWithoutColon(PDOStatement $s) { $n = 7; $s->bindValue(':n', $n); $n = 11; $s->bindValue('n', $n); $s->execute(); }
function reboundWithColon(PDOStatement $s) { $n = 7; $s->bindValue('n', $n); $n = 11; $s->bindValue(':n', $n); $s->execute(); }
// A placeholder that is not statically known may be the one bound earlier.
function reboundDynamic(PDOStatement $s, $name) { $n = 7; $s->bindValue(':n', $n); $n = 11; $s->bindValue($name, $n); $s->execute(); }
function reboundPosition(PDOStatement $s) { $n = 7; $s->bindValue(1, $n); $n = 11; $s->bindValue(1, $n); $s->execute(); }
