<?php
$users = [31 => 'Mia', 52 => 'Noor']; \uasort($users, fn($a, $b) => $a <=> $b); echo $users[31];
