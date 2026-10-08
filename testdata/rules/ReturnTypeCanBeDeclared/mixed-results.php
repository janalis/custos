<?php
// A `mixed` result (Doctrine's Query::getResult()) is no better than an
// unknown one: the method's own @return tag decides.
class Query
{
    public function getResult(): mixed
    {
        return [];
    }
}

class Item {}

class Repository
{
    public function createQuery(): Query
    {
        return new Query();
    }

    /** @return Item[] */
    public function <weak_warning descr="Declare ': array' as the return type.">findAll</weak_warning>()
    {
        return $this->createQuery()->getResult();
    }

    /** @return \Doctrine\Common\Collections\ArrayCollection */
    public function findCollection()
    {
        return $this->createQuery()->getResult();
    }

    public function untagged()
    {
        return $this->createQuery()->getResult();
    }
}
