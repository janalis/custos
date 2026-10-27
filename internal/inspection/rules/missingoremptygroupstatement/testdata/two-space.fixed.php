<?php
function bench($n) {
  for ($i = 0; $i < $n; $i++) {
    if ($i) echo $i;
  }
}
