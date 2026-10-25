<?php
final class Factory
{
    public function make($name)
    {
        return new class($name) {
            private $name;

            public function __construct($name)
            {
                $this->name = $name;
            }

            public function child()
            {
                return new class($this->name) {
                    private $name;

                    public function __construct($name)
                    {
                        $this->name = $name;
                    }

                    public function name()
                    {
                        return $this->name;
                    }
                };
            }
        };
    }
}
