<?php
interface I { public function m(); }
class C { public function m() { if (1) { return 1; } return 2; } }
