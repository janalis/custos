<?php
// presets: symfony per; set: line_length=80
class User extends Model implements
    A,
    B {}
interface I extends
    J,
    K {}
class Long extends Model implements
    AVeryLongInterfaceNameNumberOne,
    AVeryLongInterfaceNameNumberTwo {}
class C implements
    A, // why
    B {}
$x = new class (
    $a,
    $b,
) extends Foo {};
$y = new class (
    $a,
    [
        1,
    ],
) {};
