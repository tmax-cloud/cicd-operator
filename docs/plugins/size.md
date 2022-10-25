## `Size` Plugin

Size plugin measures how big the pull request is and label it to the pull request. (e.g., `size/S`, `size/XL`)

The 'size' is measured regarding the number of lines changed in the pull request. The threshold for each size tag is configurable via ConfigMap `plugin-config`.
> **Default**  
> XS : 0    ~ 10   lines  
> S  : 11   ~ 30   lines  
> M  : 31   ~ 100  lines  
> L  : 101  ~ 500  lines  
> XL : 501  ~ 1000 lines  
> XXL: 1001 ~      lines
