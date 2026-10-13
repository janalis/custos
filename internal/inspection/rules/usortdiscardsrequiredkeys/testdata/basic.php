<?php
$users = [31 => 'Mia', 52 => 'Noor']; <warning descr="Preserve record keys with uasort.">usort($users, fn($a, $b) => $a <=> $b)</warning>; echo $users[31];
