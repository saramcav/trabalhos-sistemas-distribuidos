import rpyc
import sys
import time

class Questions():
    def question1():
        if len(sys.argv) < 2:
            exit("Usage {} SERVER".format(sys.argv[0]))
        server = sys.argv[1]
        conn = rpyc.connect(server,18861)
        print(conn.root)
        print(conn.root.get_answer())
        print(conn.root.the_real_answer_though)

    def question2():
        if len(sys.argv) < 2:
            exit("Usage {} SERVER".format(sys.argv[0]))
        server = sys.argv[1]
        conn = rpyc.connect(server,18861)
        print(conn.root.get_question())
    
    def question3_4_5():
        if len(sys.argv) < 2:
            exit("Usage {} SERVER".format(sys.argv[0]))
        server = sys.argv[1]
        n = int(input('Digite o tamanho do vetor: '))
        vet = []
        for i in range(n):
            vet.append(1)
            
        start = time.time()
        conn = rpyc.connect(server,18861)
        conn._config['sync_request_timeout'] = 360
        print(f'A soma é: {conn.root.get_sum(vet)}')
        
        end = time.time()
        print(f'Tempo de execução para obter o resultado no cliente: {end-start}')


        


if __name__ == "__main__":
    start = time.time()

    #Questions.question1()
    #Questions.question2()
    Questions.question3_4_5()
    
    end = time.time()
    print(f'Tempo de execução total: {end-start}')