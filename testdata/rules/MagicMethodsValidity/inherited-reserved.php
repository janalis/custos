<?php
interface Gatekeeper
{
    public function __allows();
}
abstract class BaseGate implements Gatekeeper
{
    abstract public function __realm(): string;
}
class OfficeGate extends BaseGate
{
    public function __allows() { return true; }
    public function __realm(): string { return 'office'; }
    public function <error descr="The '__' prefix is reserved for magic methods.">__extra</error>() {}
}
