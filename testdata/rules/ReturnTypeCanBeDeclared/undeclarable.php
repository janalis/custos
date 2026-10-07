<?php

final class Shutdown
{
    private $data;

    public function stop()
    {
        echo 'bye';

        return exit();
    }

    /**
     * Returns the rows.
     *
     * @return {array}
     */
    public function rows()
    {
        return $this->data['rows'];
    }

    /** @return self::KIND_* */
    public function kind()
    {
        return $this->data['kind'];
    }

    public function <weak_warning descr="Declare ': array' as the return type.">all</weak_warning>()
    {
        return func_get_args();
    }
}
