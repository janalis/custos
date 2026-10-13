<?php
function otherName(PDOStatement $s) { $n = 7; $s->bindValue(':n', $n); $n = 11; $s->bindValue('m', $n); <warning descr="Rebind the changed value before execution.">$s->execute()</warning>; }
function otherPosition(PDOStatement $s) { $n = 7; $s->bindValue(1, $n); $n = 11; $s->bindValue(2, $n); <warning descr="Rebind the changed value before execution.">$s->execute()</warning>; }
function positionAgainstName(PDOStatement $s) { $n = 7; $s->bindValue(1, $n); $n = 11; $s->bindValue(':n', $n); <warning descr="Rebind the changed value before execution.">$s->execute()</warning>; }
