<?php
abstract class CacheBase
{
    abstract protected function _get($key);
}
class CacheFile extends CacheBase
{
    protected function _get($key) { return null; }
    protected function <error descr="'_set' is not magic; did you mean '__set'?">_set</error>($key, $value) {}
}
$o = new class {
    public function <error descr="'_call' is not magic; did you mean '__call'?">_call</error>($m) {}
};
