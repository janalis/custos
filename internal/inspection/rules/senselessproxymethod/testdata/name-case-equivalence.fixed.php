<?php
class Mode { const FAST = 1; }
class Runner {
    public function run($mode = Mode::FAST) { return $mode; }
}
class FastRunner extends Runner {
    }
