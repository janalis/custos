<?php
$r = <warning descr="Compare directly: ''admin' !== $role'.">!in_array($role, ['admin'])</warning>;
$s = <warning descr="Compare directly: ''x' === $role'.">in_array($role, ['x'], $flag)</warning>;
