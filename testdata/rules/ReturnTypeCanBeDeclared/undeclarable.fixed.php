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

    public function all(): array
    {
        return func_get_args();
    }
}
