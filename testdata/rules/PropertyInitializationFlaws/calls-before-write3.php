<?php
class Loader
{
    private $paths = [];

    public function __construct(array $paths)
    {
        $this->init(); // may read the default
        $this->paths = $paths;
    }

    private function init(): void {}
}
