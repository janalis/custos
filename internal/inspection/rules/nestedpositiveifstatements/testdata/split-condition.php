<?php
function split_condition($a, $b, $c, $d) {
    if ($a
      && $b) {

      <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($c && $d) {
        go();
      }
    }
}
