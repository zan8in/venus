## venus
subdomain scanner

## Features
- Get subdomain from **FOFA** 
- **Dictionary blasting** 

## Example

Domain Input
```
venus -t yourdomain.com
```

Multiple Domain Input
```
venus -t yourdomain.com,yourdomain2.com,yourdomain3.com
```

List of Domains Input
```
venus -T domain_list.txt
```

```
$ cat domain_list.txt

yourdomain.com
yourdomain2.com
yourdomain3.com
...
```

Output files
```
venus -T domain_list.txt -o r.txt
```

## Build

Go Version
```
go >= 1.19
```

build
```
go build cmd/venus
```

