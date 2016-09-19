# -*- coding: utf-8 -*-
"""

@author: navid
"""
import time
import struct
import os
import random
import hashlib
from threading import  Thread
import subprocess
BLOCK_SIZE=512

class Record:
    # record=bytearray()
    # checksum=bytearray()
    # timestamp=bytearray()
    # block=bytearray()
    # worker_id=bytearray()
    # op_cnt=bytearray()
    # seed=bytearray()
    # marker=bytearray()

    def __init__(self,address,worker_id,op_cnt,seed,):
        # self.set_time()
        self.address=struct.pack('i',address)
        self.worker_id=struct.pack('i',worker_id)
        self.op_cnt=struct.pack('i',op_cnt)
        self.seed=struct.pack('i',seed)
        self.set_marker()
        self.set_time()
        tmp=self.address+self.worker_id+self.op_cnt+self.seed+self.marker
	tmp2=str(address)+','+str(worker_id)+','+str(op_cnt)+','+str(seed)
        record=bytearray()
        #for index in range(1,129):
        #    record+=tmp
	record=tmp2
        self.record=tmp2
        print "block"
        print address
        print "worker"
        print worker_id
        print "opcnt"
        print op_cnt
        print "seed"
        print seed
	print "record"
	print tmp2
	


        # print self.record
        # print len(self.marker)


    def set_time(self):
        # global timestamp
        tmp_time=time.clock()

        self.timestamp=struct.pack('f', tmp_time)
        print "time"
        print tmp_time


    def set_marker(self):
        # global marker
        self.marker=b'\x23\x23\x23\x23\x23\x23\x23\x23\x23\x23\x23\x23'
        # print marker


def s_thread_seq(): #single threaded sequential writes
    with open('/home/navid/navid/IdeaProjects/SSD/test1', 'r+b') as f:

        t1=time.clock()
        t2=t1
        address=0
        while address<100:

            # print "1"
            #
            # a=time.clock()


            r=Record(address,1,address,1)
            # r.set_time()
            # print opcnt
            # print "record after"
            # print r.record
            # print "time after"
            # print r.timestamp
            # print "new len"
            # print len(r.record)

            f.write(r.record)
            address+=1

            t2=time.clock()
            # b=time.clock()
            # print b-a


        f.close()

def s_thread_rand(seed):
    print "11"
    with open('/dev/sdb', 'r+b') as f:

        for opcnt in range(1, 2):
            rand= random.random()
            address=(int)(rand*100)
            print "address"
            print (int)(rand*100)
            f.seek(address)
            r=Record(address,1,opcnt,seed)
            # print r.record
            #f.write(r.record)
	    command=['/home/navid/scsi_test/./scsi_test' , "%s"%(r.record)]
	    subprocess.call(command)	
            print "this is record%s"%(command)
	    #print r.record	
	    #print a

        f.close()

def testing():
    #m = hashlib.md5()
    #m1= hashlib.md5()	
    with open('chartest', 'r+b') as f:
        line=bytearray()
	counter_address=0
	with open('/dev/sdb', 'r+b') as SSD:
		line=f.read(512)
		while(len(line)!=0):
			
			print "line"
			print(len(line[4:8]))

			tmp_time= struct.unpack('f',line[0:4])
			time=tmp_time[0]
			print time

			tmp_block=struct.unpack('i',line[4:8])
			block=tmp_block[0]
			print block

			tmp_address=struct.unpack('i',line[8:12])
			address=tmp_address[0]        
			print address

			tmp_worker_id=struct.unpack('i',line[12:16])
			worker_id=tmp_worker_id[0]
			print worker_id

			tmp_opcnt=struct.unpack('i',line[16:20])
			opcnt=tmp_opcnt[0]        
			print opcnt

			tmp_seed=struct.unpack('i',line[20:24])
			seed=tmp_seed[0]        
			print seed

			ssd_address=address *BLOCK_SIZE
			print ssd_address
			SSD.seek(ssd_address)
			s_line=SSD.read(512)

			m=hashlib.md5()
			m.update(line)
			checksum=m.digest()

			m1=hashlib.md5()
			m1.update(s_line)
			s_checksum=m1.digest()
			if(s_checksum!=checksum):
				print "hello"
				print counter_address
				index=0
				header_checksum=s_line[24:40]
				m=hashlib.md5()
				m.update(s_line[4:24])
				c_header_checksum=m.digest()
				while(c_header_checksum==header_checksum):
					index=index+1
					header_checksum=s_line[24+index*40:40+index*40]
					m=hashlib.md5()
					m.update(s_line[4+index*40:24+index*40])
					c_header_checksum=m.digest()

				with open("errorlog" ,"a") as error:
					error_message="message: ADDRESS => %s Number of Correct Headers: %s\n"%(address,index)
					error.write(error_message);
					error.close()
					

			counter_address=counter_address+1
			f.seek(counter_address*512)

			line=f.read(512)
			
	
			

	
	
		
		
			
		
		SSD.close()
		

	

def concurrent_rand():
    # testing()
    # thread.start_new_thread(testing,())
        # thread.start_new_thread(s_thread_rand,)
    # print "why"
        # s_thread_rand()
    t1=Thread(target=s_thread_rand(2),args=( ))
    t2=Thread(target=s_thread_rand(1),args=( ))
    t1.start()
    t2.start()


#concurrent_rand()
testing()









