 server {
        listen 443;
        server_name nr-supportlanding.s3.amazonaws.com;
        root /usr/share/nginx/upload;
        ssl on;
        ssl_certificate /etc/nginx/ssl/nr-supportlanding.s3.amazonaws.com/aws.crt;
        ssl_certificate_key /etc/nginx/ssl/nr-supportlanding.s3.amazonaws.com/aws.pem;
        location /staging {
            client_max_body_size   10000m;
            dav_methods            PUT ;
            client_body_temp_path  upload/client_tmp;
            create_full_put_path   on;
            dav_access             group:rw  all:r;
            return                 200;
        }
}
