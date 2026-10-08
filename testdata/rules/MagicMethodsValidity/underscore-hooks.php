<?php
abstract class Model
{
    public function __construct()
    {
        $this->_construct();
    }

    protected function _construct()
    {
    }
}

class Decorator
{
    public function __invoke(...$args)
    {
        return $this->_call($args);
    }

    protected function _call(array $args)
    {
        return $args;
    }
}

class Builder
{
    protected $data = [];

    protected function <error descr="'_set' is not magic; did you mean '__set'?">_set</error>($key, $value)
    {
        $this->data[$key] = $value;
    }
}

class Child extends Model
{
    protected function <error descr="'_clone' is not magic; did you mean '__clone'?">_clone</error>()
    {
    }
}
