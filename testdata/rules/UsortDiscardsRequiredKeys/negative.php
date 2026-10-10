<?php
$users = ['Mia', 'Noor']; usort($users, fn($a, $b) => $a <=> $b); echo $users[0];

$a=['k'=>1];usort($a,$cmp);$a['k']=9;
