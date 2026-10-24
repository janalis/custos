<?php
// The companion may come from an ancestor that is not indexed, or from the
// class using the trait.
class PublicationDates extends \Drupal\Core\Field\ChangedFieldItemList
{
    public function __get($name)
    {
        return parent::__get($name);
    }
}

trait StorageTrait
{
    public function __get($name)
    {
        return $this->$name ?? null;
    }
}

class Storage
{
    use StorageTrait;

    public function __set($name, $value) {}

    public function __isset($name)
    {
        return false;
    }
}

class Known
{
    public function <error descr="__get needs a companion __set method.">__get</error>($name)
    {
        return null;
    }
}
