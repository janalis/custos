<?php
if (<warning descr="Read the new mask with a separate umask call.">umask(0077) !== 0077</warning>) { throw new RuntimeException(); }
