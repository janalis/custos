<?php
class Repo
{
    protected function load($id)
    {
        $row = ['id' => $id];
        return strtoupper(json_encode($row));
    }

    protected function save($row, $flag)
    {
        $this->log($row);
    }

    protected function fetch(...$ids)
    {
        return $this->log($ids);
    }

    /** @return int */
    protected function total()
    {
        echo 1;
    }

    protected function count()
    {
        return 1;
    }

    protected function secret()
    {
        return $this->hidden();
    }

    private function hidden()
    {
        return 0;
    }

    public function log($x)
    {
        echo $x;
    }
}

class MiddleRepo extends Repo {}

class UserRepo extends MiddleRepo
{
    /**
     * Copied from Repo.
     */
    protected function <weak_warning descr="Method 'load' duplicates the inherited implementation; remove it.">load</weak_warning>($key)
    {
        // same code, different comments
        $row = ['id' => $id];

        /** stray doc block */
        return strtoupper(json_encode( $row ));
    }

    public function <weak_warning descr="Method 'save' duplicates the inherited implementation; delegate to parent::save() instead.">save</weak_warning>($row, $flag)
    {
        $this->log($row);
    }

    public function <weak_warning descr="Method 'fetch' duplicates the inherited implementation; delegate to parent::fetch() instead.">fetch</weak_warning>(...$ids)
    {
        return $this->log($ids);
    }

    /** @return int */
    public function <weak_warning descr="Method 'total' duplicates the inherited implementation; delegate to parent::total() instead.">total</weak_warning>()
    {
        echo 1;
    }

    protected function count()
    {
        return 2;
    }

    protected function secret()
    {
        return $this->hidden();
    }
}
