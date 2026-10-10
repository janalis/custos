<?php
// @custos-ignore GdColorZeroRejectedAsFailure

if(!$color=imagecolorallocate($i,0,0,0)){throw new RuntimeException();}function color($i){if(imagecolorallocatealpha($i,0,0,0,1)==false){return false;}}
