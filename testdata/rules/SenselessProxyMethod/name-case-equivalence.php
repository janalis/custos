<?php
class Mode { const FAST = 1; }
class Runner {
    public function run($mode = Mode::FAST) { return $mode; }
}
class FastRunner extends Runner {
    public function <weak_warning descr="Method 'run' only forwards to its parent; remove it.">run</weak_warning>($mode = MODE::FAST) { return parent::run($mode); }
}
