<?php
if (!<warning descr="Distinguish regex failure from a successful non-match.">preg_match($pattern, $text)</warning>) { return false; }
