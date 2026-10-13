<?php
interface A {}
interface B extends A {}
interface C extends B {}
interface D extends C {}
$x = new class extends Exception {};
// A cycle (a fatal error in PHP) stops the walk.
class Loop1 extends Loop3 {}
class Loop2 extends Loop1 {}
class Loop3 extends Loop2 {}
