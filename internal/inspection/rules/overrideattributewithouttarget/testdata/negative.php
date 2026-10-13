<?php
interface Starter { public function start(); } class Worker implements Starter { #[\Override] public function start() {} }
trait Starts { #[\Override] public function start() {} }
class Base { public function start() {} } class Runner extends Base { use Starts; }
class Label { #[\Override] public function __toString(): string { return 'label'; } }
