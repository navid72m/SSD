

#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <time.h>
#include <string.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <openssl/md5.h>
#include "scsiexe.h"

#define BLOCK_LENGTH_BYTE 512
#define HEADER_LENGTH_BYTE 40

#define BLOCK_SIZE BLOCK_LENGTH_BYTE/4
#define HEADER_SIZE HEADER_LENGTH_BYTE/4 
unsigned char sense_buffer[SENSE_LEN];
uint32_t data_buffer[BLOCK_SIZE];

void show_hdr_outputs(struct sg_io_hdr * hdr) {
	printf("status:%d\n", hdr->status);
	printf("masked_status:%d\n", hdr->masked_status);
	printf("msg_status:%d\n", hdr->msg_status);
	printf("sb_len_wr:%d\n", hdr->sb_len_wr);
	printf("host_status:%d\n", hdr->host_status);
	printf("driver_status:%d\n", hdr->driver_status);
	printf("resid:%d\n", hdr->resid);
	printf("duration:%d\n", hdr->duration);
	printf("info:%d\n", hdr->info);
}

void show_sense_buffer(struct sg_io_hdr * hdr) {
	unsigned char * buffer = hdr->sbp;
	int i;
	for (i=0; i<hdr->mx_sb_len; ++i) {
		putchar(buffer[i]);
	}
}

void show_vendor(struct sg_io_hdr * hdr) {
	unsigned char * buffer = hdr->dxferp;
	int i;
	printf("vendor id:");
	for (i=8; i<16; ++i) {
		putchar(buffer[i]);
	}
	putchar('\n');
}

void show_product(struct sg_io_hdr * hdr) {
	unsigned char * buffer = hdr->dxferp;
	int i;
	printf("product id:");
	for (i=16; i<32; ++i) {
		putchar(buffer[i]);
	}
	putchar('\n');
}

void show_product_rev(struct sg_io_hdr * hdr) {
	unsigned char * buffer = hdr->dxferp;
	int i;
	printf("product ver:");
	for (i=32; i<36; ++i) {
		putchar(buffer[i]);
	}
	putchar('\n');
}

void test_execute_Inquiry(char * path, int evpd, int page_code) {
	
	printf("starting\n");
	struct sg_io_hdr * p_hdr = init_io_hdr();

//**************Getting command from Python and Parsing it****************//
	//int i=0;
	//printf(path);
	char * token;
	uint32_t commands[4];
	int ADDRESS;
	token=strtok(path,",");
	commands[0]=atoi(token);
	printf("%s\n",token);
	int index=1;

	while(token !=NULL && index<4){
		token=strtok(NULL,",");
		printf("%s",token);
		commands[index]=atoi(token);
		//printf("%c\n",commands[index]);
		index++;
		printf("%d",index);
	}
	ADDRESS=(int)commands[0];
//************************************************************************//

//***********************Generating the Record****************************//

	uint32_t Record[BLOCK_SIZE]={0};
	uint32_t Block_tmp[6]={0};
	int p=0,t=0,block_num=0;
	unsigned char digest[16];
	//const char* string = "Hello World";
	MD5_CTX context;
	clock_t t1;

	for(block_num=0 ; block_num < 12 ; block_num++){
		
		Block_tmp[0]=block_num;
		Block_tmp[1]=commands[0];
		Block_tmp[2]=commands[1];
		Block_tmp[3]=commands[2];
		Block_tmp[4]=commands[3];
		

		Record[(block_num*10)+1]=(uint32_t) block_num;
		Record[(block_num*10)+2]=commands[0];
		Record[(block_num*10)+3]=commands[1];
		Record[(block_num*10)+4]=commands[2];
		Record[(block_num*10)+5]=commands[3];

		MD5_Init(&context);
		MD5_Update(&context,Block_tmp ,20);
		MD5_Final(digest, &context);

		for(t=0 ; t<4 ; t++){
		Record[(block_num*10)+6+t]=Record[(block_num*10)+6+t] | (uint32_t)digest[(t*4)+3] << 24|(uint32_t)digest[(t*4)+2] << 16|(uint32_t)digest[(t*4)+1] <<8|(uint32_t)digest[(t*4)] ;
                                                                                                                                       								 								
		}
		Record[(block_num*10)+10]=(uint32_t)0x23 <<24 | (uint32_t)0x23 <<16 | (uint32_t)0x23 <<8 | (uint32_t)0x23 ; //padding for header

		
	}
	
	for (p=0 ; p<7 ; p++){
	Record[121+p]=(uint32_t)0x23 <<24 | (uint32_t)0x23 <<16 | (uint32_t)0x23 <<8 | (uint32_t)0x23 ;    //padding for Record
	}
	
	
	printf("BLOCK_SIZE: %d\n",BLOCK_SIZE);
	FILE *fp;
	//char str[]="Hello World!";
	fp=fopen("chartest" , "w");
	if(fp!=NULL){
	printf("file exist");
	}
	
	
	
	union FloatOrUInt
    	{
        	float asFloat;
        	unsigned int asUInt;
    	} fu;

 	t1=clock();
    	fu.asFloat = t1;
 
    	uint32_t uint;
 
    	uint = fu.asUInt;
 
    	Record[0] = (uint32_t)uint;
 
	
	fwrite(Record , 1 , sizeof(Record), fp);
	fclose(fp);
	printf("time%f",(float)t1);
	//int f=open("chartest",O_CREAT | O_RDWR);
	
	
	
	
	
	printf("finished");
	
	printf("initializing");
	set_xfer_data(p_hdr, Record,512);
	set_sense_data(p_hdr, sense_buffer, SENSE_LEN);

	int status = 0;
	printf("%s",path);
	int fd = open("/dev/sdb", O_RDWR);
	if (fd>0) {
		status = execute_Inquiry(fd, ADDRESS, p_hdr);
		printf("the return status is %d\n", status);
		if (status!=0) {
			show_sense_buffer(p_hdr);
		} else{
			//show_vendor(p_hdr);
			//show_product(p_hdr);
			//show_product_rev(p_hdr);
		}
	} else {
		printf("failed to open sg file %s\n", path);
	}
	close(fd);
	destroy_io_hdr(p_hdr);
}

int main(int argc, char * argv[]) {
	printf("before starting");
	test_execute_Inquiry(argv[1], 0, 0);
	return EXIT_SUCCESS;
}
