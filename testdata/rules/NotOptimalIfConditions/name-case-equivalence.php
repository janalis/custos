<?php
interface Shape {}
class Circle implements Shape {}
class Ring extends Circle {}
class Repo { public static $s; public static function current() { return null; } }

if (Repo::$s instanceof Shape || <warning descr="Redundant instanceof: another check on the same value already covers this type.">repo::$s instanceof Ring</warning>) {}
if (<weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">REPO::$s != null</weak_warning> && Repo::$s instanceof Ring) {}
if (Repo::$s instanceof Shape || Repo::$S instanceof Ring) {}
