<?php
$r = <warning descr="Compare directly: ''admin' !== $role'.">!in_array($role, ['admin'])</warning>;
$s = in_array($role, ['x'], $flag);
