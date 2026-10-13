<?php
interface Shape {}
class Square implements Shape {}
class Stamp {}
class Patched extends MissingBase {}

function paint(Shape $shape, Shape $other, Shape $third) {
    $shape = new UnknownShape();
    $other = new Patched();
    $third = <warning descr="Assigning a value of type \Stamp does not match the parameter's declared type.">new Stamp()</warning>;
    $third = new Square();
    return [$shape, $other, $third];
}
