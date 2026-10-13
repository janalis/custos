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
    public function save($row, $flag)
    {
        parent::save($row, $flag);
    }

    public function fetch(...$ids)
    {
        return parent::fetch(...$ids);
    }

    /** @return int */
    public function total()
    {
        return parent::total();
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
