<?php
interface Runs { public function run(); }
abstract class Job implements Runs {
    abstract public function step();
    public function go() { self::step(); static::run(); \Job::go2(); }
    public function go2() {}
}
class Logger { public function log() {} }
class Unrelated {
    public function log() {}
    public function work() { Logger::log(); make()::log(); Missing::thing(); self::nope(); }
}
function make(): Logger { return new Logger(); }
