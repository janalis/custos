<?php
if (<warning descr="Reject permission lookup failure before checking bits.">(fileperms($path) & 0002) === 0</warning>) { return true; }
